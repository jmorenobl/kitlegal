# Quickstart: comprobar H45

Desde la raíz del repositorio, en la rama del hito ya implementada y con la suite activada. Requisitos: el Go de
`go.mod` y las herramientas de `make check-tools`. Ninguna orden toca la red salvo `make ci` (su paso `vuln`) ni abre una
sesión con modelo. Efectos: `go test` escribe solo en la caché de Go y en directorios temporales; `make ci` deja además
`coverage.out` y `coverage-integration.out` en la raíz, que `.gitignore` ignora (`/*.out`). Ninguna orden cambia el árbol
de trabajo, el índice ni el historial de git.

## 1. Lo de hoy, igual (US1 · FR-008, FR-026 · SC-001)

```bash
make skills-check
git diff --stat main -- skills/
```

Esperado: `skills-check` en verde (las referencias, idénticas a lo que se genera desde `data/`), y el diff solo en
`skills/normativa/SKILL.md`, `skills/normativa/references/con_herramientas.md` y `skills/plazos-y-territorio/SKILL.md`.

```bash
git diff main -- skills/plazos-y-territorio/references/ skills/normativa/references/normas.md
go test -count=1 -run '^(TestCasosDePlazos|TestAppletPlazos)$' ./internal/app/
go test -count=1 -run '^TestEntregaDelHito$/^(h9-plazos-.*|h36-plazos-de-hoy)$' ./internal/app/
```

Esperado: el primero, vacío; los otros dos, en verde.

## 2. El molde con la vertical de ensayo (US2 · FR-002, FR-005, FR-006, FR-019, FR-020 · SC-002)

```bash
go test -count=1 -v -run '^TestVerticalDeEnsayo$' ./internal/skills/
```

Esperado: `PASS`, con sus subtests `mutante` y `sin-la-region`.

## 3. Los defectos nuevos (US3 · FR-003, FR-004 · SC-003)

```bash
go test -count=1 -v -run '^TestRegiones$/^defectos$/^(tipo-sin-paquete|tipo-de-dos-paquetes|computo-de-otro-paquete)$' ./internal/skills/
```

Esperado: los tres subtests en `PASS`.

## 4. Un paquete más no rompe la lista (US4 · FR-016 · SC-004)

```bash
go test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/
```

Esperado: `PASS`. Tarda unos segundos más que otro guion: copia el módulo a su `$WORK` (unos 150 MB, que testscript borra
al acabar) y corre dos `go test` en la copia.

## 5. `normativa` hace sitio y el control lo admite (US5 · FR-009 a FR-013 · SC-008)

```bash
tail -n 4 skills/normativa/references/con_herramientas.md
grep -c 'En PowerShell 5.1:' skills/normativa/SKILL.md
go test -count=1 -run '^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/
```

Esperado: las cuatro líneas de §C.2 del spec; `0` (y `grep` sale con 1, porque no encuentra nada); los tests, en verde.

## 6. El laboratorio y la descripción (US6, US9 · FR-010, FR-014, FR-023 · SC-007, SC-009)

```bash
tr -d '\n' < laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt | wc -m
go test -count=1 -run '^(TestSkillsDelProductoDesdeControl|TestDescripcionesCortas|TestDescripcionesDelProducto)$' ./internal/evals/
go test -count=1 -v -run '^TestEntregaDelHito$/^(h41-4-skills-instaladas|h45-skills-instaladas)$' ./internal/app/
```

Esperado: `429`; los tests y los dos guiones, en verde.

## 7. Las grabaciones y el trabajo de evals (US7, US8 · FR-015, FR-021, FR-022 · SC-005, SC-006)

```bash
go test -count=1 -run '^(TestGrabacionesSinSolape|TestManifiestoDeGrabaciones|TestDefinicionDelJob|TestPlanEnDosModos)$' ./internal/evals/
grep -n -A1 'skill: plazos-y-territorio' .github/workflows/evals.yml
```

Esperado: en verde; en el trabajo `evals`, `concurrencia: 3`.

## 8. Todo

```bash
make ci
```

Esperado: `ci: todos los controles en verde` (SC-011).
