# Quickstart: validar H31.1

Desde la raíz del repositorio, en la rama del hito y con la entrega hecha. Sin red ni modelo. Todos los escenarios
leen y no escriben nada en el árbol, el índice ni el historial, salvo el 12, que deja `coverage.out` y
`coverage-integration.out`, ignorados por `.gitignore` (`/coverage.*`). Los que se pueden ejecutar hoy, antes de la
entrega, lo dicen con su resultado de hoy.

## 1. El cuerpo cabe (FR-005, SC-002)

```bash
tail -n +16 skills/boe-legislacion/SKILL.md | wc -m
```

Con la entrega: `10995`. Hoy: `10992`. La skill medida, `tail -n +16 evidencias/sondeos-20261026/brazo-W/skill/SKILL.md | wc -m`: `12081`.

## 2. El `SKILL.md` es el del contrato (FR-001 a FR-010)

```bash
cmp skills/boe-legislacion/SKILL.md specs/029-h31-1-boe-legislacion-habla/contracts/skill-md-v0.3.1.md && echo igual
```

Con la entrega: `igual`. La tabla de textos de `contracts/skill-boe-legislacion-v0-3-1.md` §2 dice cada diferencia con
la skill medida: `diff evidencias/sondeos-20261026/brazo-W/skill/SKILL.md skills/boe-legislacion/SKILL.md`.

## 3. Las referencias son las de la skill medida (FR-001, FR-007)

```bash
diff -r evidencias/sondeos-20261026/brazo-W/skill/references skills/boe-legislacion/references && echo iguales
```

Con la entrega: `iguales`. Hoy: difieren `buscar_la_norma.md` y `redaccion_modificada.md`.

## 4. La suite de aceptación (SC-003)

```bash
go test -count=1 -v -run 'TestEntregaDelHito/h31-1-' ./internal/app/ 2>&1 | grep -E '^\s*--- '
```

Con la suite activada: `--- PASS` para `h31-1-responder`, `h31-1-lee-lo-que-nombra` y `h31-1-sitio-en-el-cuerpo`.

## 5. Las skills (FR-003, FR-004, FR-007, FR-008; SC-001)

```bash
make skills-check
```

Sale con 0, con `ok` en `internal/app`, `internal/skills` e `internal/evals`.

## 6. Las lecturas y sus grabaciones (FR-011 a FR-013; SC-004, SC-008)

```bash
go test -count=1 -v -run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestLecturaDeLaEvalSinGrabar)$' ./internal/evals/ 2>&1 | grep -E -- '--- (PASS|FAIL): Test[A-Za-z]+(/(lecturas|grabado))?( |$)'
```

Con la entrega: `PASS` en `TestEvalsDelRepositorio`, `TestEvalsDelRepositorio/lecturas`, `TestEvalsDelRepositorio/grabado`,
`TestCacheDeLasSesiones`, `TestCacheDeLasSesiones/lecturas` y `TestLecturaDeLaEvalSinGrabar`. Las grabaciones nuevas:

```bash
ls internal/evals/testdata/boe.legislacion-consolidada/ | grep -c -E '_(BOE-A-2006-20764_texto_bloque_(a7|a19|a20|dadecimoctava)|BOE-A-2004-4214_texto_bloque_(a60|a92|a100|a104))\.json$'
```

Con la entrega: `8`. Hoy: `0`.

## 7. El tope del trabajo de evals (FR-015, SC-005)

```bash
go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/
grep -n 'timeout-minutes: 352' .github/workflows/evals.yml
```

`ok`, y la línea 276. Hoy, igual: el hito no lo cambia (research D8).

## 8. Lo que la rama añade a la skill es genérico (FR-009)

```bash
git diff main -- skills/boe-legislacion/ | grep '^+' | grep -v '^+++' | grep -c -e BOE-A-2004-4214 -e BOE-A-2006-20764 -e BOE-A-2013-12887 -e 'Ley 35/2006' -e 'Real Decreto Legislativo 2/2004' -e 'Ley 19/2013' -e IRPF -e TRLRHL -e LTAIBG
```

`0` (y código 1, el de `grep -c` sin coincidencias). Hoy, igual: la rama aún no toca la skill.

## 9. El diff de `evals/` solo añade las nueve lecturas (FR-011)

```bash
git diff main --numstat -- evals/
git diff main -- evals/ | grep -c '^-[^-]'
```

Con la entrega: tres líneas, de `14-…`, `15-…` y `17-…`, con `0` borradas en cada una (12 añadidas en la 14 y en la 15,
4 en la 17: tres por lectura y la clave `lecturas:` de la 17), y `0`.

## 10. Las URL de §1 de `docs/SKILLS.md` (FR-014, SC-013)

```bash
git diff main -- docs/SKILLS.md | grep '^+' | grep -o 'https://[^)> ]*' | sort -u
grep -n 'https://agentskills.io/' web/src/components/Pie.astro
```

Con la entrega: solo `https://agentskills.io/`, y la línea 30 de `Pie.astro`. Las cifras del párrafo se cotejan con
`sed -n '57,59p' evidencias/sondeos-20261025/LEEME.md`.

## 11. La entrada de `CHANGELOG.md` (FR-016)

```bash
awk '/^## \[Unreleased\]/ {u = 1; next} /^## \[/ {u = 0} u' CHANGELOG.md | grep -c 'boe-legislacion` v0.3.1'
```

Con la entrega: `1`. Hoy: `0`.

## 12. `make ci` (SC-001)

```bash
make ci
```

Termina con `ci: todos los controles en verde`. Escribe `coverage.out` y `coverage-integration.out`, ignorados.

## Lo que no se valida aquí

El cierre (SC-006 a SC-011): el job de evals sobre la cabeza de la propuesta, que lanza el workflow. Y el sondeo de la
eval 15 con la skill de la propuesta (SC-012), que lanza una persona antes de fusionar, con el arnés de
`evidencias/sondeos-20261025/` y `evidencias/sondeos-20261026/`.
