# Quickstart · H32 · cómo comprobar la entrega

Desde la raíz del repositorio, con la rama del hito implementada. Ningún escenario escribe en el árbol de trabajo, el
índice ni el historial: los tests escriben en sus directorios temporales y en la caché de Go, y el binario de prueba
y sus ámbitos van a `$TMPDIR`. Ninguno abre una sesión con modelo ni toca la red.

## 1. Las skills y sus controles (FR-005 a FR-013, FR-018, FR-023; SC-001 a SC-004)

```sh
make skills-check
go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/
go test -count=1 -run '^(TestSkillsEmpotradas|TestConfiguracionDeLaRelease)$' .
```

Esperado: los tres en verde. El segundo corre, con sus mutantes, `forma-de-cita` y `formas-compartidas` (los controles
nuevos de FR-010 y FR-013) y los de hoy (`normas-nombradas`, `sin-instrucciones-de-evals`, presupuestos, regiones). El
tercero, los dos tests de la raíz que nombran las skills: `skillsDeHoy` con las cuatro de la entrega y la orden del
humo de `skills install` con `normativa` (research D19).

Las cuatro skills, y el cuerpo de cada una medido como el gate:

```sh
ls skills/
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/normativa/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/plazos-y-territorio/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
```

Esperado: `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio`; cuerpos de 10 858 y 10 896
caracteres o menos ([contracts/textos.md](contracts/textos.md)).

## 2. El paquete de lo administrativo (FR-012, FR-019, FR-023)

```sh
grep -n 'paquetes' skills/plazos-y-territorio/SKILL.md
grep -n 'aplicación de los tributos' skills/plazos-y-territorio/references/administrativo.md
```

Esperado: las dos marcas con la línea `- Si la pregunta es de lo administrativo, lee [administrativo](references/administrativo.md).`
entre ellas, y la fila de los tributos con su línea `⚠ NO CUBIERTO:` entera, la de
[contracts/paquetes.md](contracts/paquetes.md) §3.

## 3. La vertical de ensayo (FR-024; SC-005)

```sh
go test -count=1 -run '^TestVerticalDeEnsayo$' -v ./internal/skills/
```

Esperado: `--- PASS: TestVerticalDeEnsayo`, con su subtest del mutante en verde (el mutante da su defecto).
`git status --porcelain skills/ data/` sigue vacío.

## 4. La instalación con un nombre retirado (FR-025 a FR-031; SC-006)

```sh
go test -count=1 -run '^TestSkillsRetiradas$' ./internal/app/
go test -count=1 -run '^TestEntregaDelHito$/^h32-retiradas$' -v ./internal/app/
```

Y a mano, con un binario construido en `$TMPDIR` y un ámbito vacío:

```sh
go build -o "$TMPDIR/kitlegal-h32" ./cmd/kitlegal
mkdir -p "$TMPDIR/h32-ambito"
"$TMPDIR/kitlegal-h32" skills install boe-legislacion --dir "$TMPDIR/h32-ambito"; echo "salida: $?"
"$TMPDIR/kitlegal-h32" skills install --dir "$TMPDIR/h32-ambito" --json
```

Esperado: la primera, salida 2 y el mensaje con las skills disponibles (`consulta-juridica, jurisprudencia, normativa,
plazos-y-territorio`); la segunda, sale con 0 y `data` lista las cuatro con `estado` `instalada`.

## 5. La suite de aceptación (SC-017, SC-018)

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h32-' -v ./internal/app/
```

Esperado: `PASS` de los seis guiones `h32-*` de [contracts/aceptacion.md](contracts/aceptacion.md). La búsqueda de
FR-032, sola:

```sh
git grep -n -E \
  -e '(^|[^A-Za-z0-9_./-])(skills|evals)/(boe-legislacion|legal-core)([^A-Za-z0-9-]|$)' \
  -e '(^|[^A-Za-z0-9_./-])\.agents/skills/(boe-legislacion|legal-core)([^A-Za-z0-9-]|$)' \
  -e '`(boe-legislacion|legal-core)`' \
  -e '(^|[^A-Za-z0-9_])[Ss]kills?[[:space:]]+`?(boe-legislacion|legal-core)([^A-Za-z0-9-]|$)' -- \
  skills evals internal cmd .github/workflows Makefile README.md web/src/pages \
  ':(exclude)evals/*/juez/*' ':(exclude)internal/evals/testdata' \
  ':(exclude)internal/app/retiradas.go' ':(exclude)internal/app/retiradas_test.go' \
  ':(exclude)internal/app/testdata/script/h32-retiradas.txtar' \
  ':(exclude)internal/app/testdata/script/h32-nombres-retirados.txtar' | grep -v 'alt="'
```

Esperado: ninguna línea (sale con 1).

## 6. Las evals, el job y la elección (FR-014, FR-033 a FR-039; SC-008)

```sh
go test -count=1 -run '^(TestEvalsDelRepositorio|TestDefinicionDelJob|TestEleccionDeLosMutantes|TestLimiteDeLaTasa|TestUmbralesDeLasSkills|TestCacheDeLasSesiones|TestJuzgar|TestJuzgarPlazos|TestIdentificadoresDeLasNormas)$' ./internal/evals/
git diff --find-renames --name-status c991808 -- evals/
```

Esperado: los tests en verde; el diff, con cada fichero de `evals/boe-legislacion/` y `evals/legal-core/` como `R…`
hacia `evals/normativa/` y `evals/plazos-y-territorio/`, y `A` solo para los dos de `evals/consulta-juridica/`.

## 7. Todo

```sh
make ci
```

Esperado: «ci: todos los controles en verde», y `git status --porcelain` igual que antes de ejecutarlo: lo que los
objetivos escriben fuera de los temporales y de la caché de Go es lo que `.gitignore` ya ignora (`/bin/`, `/dist/`,
`/*.out`, `/coverage.*`).

## Lo que no se comprueba aquí

El job de evals del cierre (SC-009 a SC-015) lo mide el workflow sobre la cabeza de la rama, y el sondeo de 280
sesiones (SC-016) lo lanza una persona antes de fusionar, con el arnés de `evidencias/sondeos-20261028/`.
