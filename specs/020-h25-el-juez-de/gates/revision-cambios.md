# Qué juzga la ronda 3 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 2, sobre `4a127c5` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `4a127c5..HEAD` (`0dc549c`):

- `0dc549c` fix(H25): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/doc.go`, `internal/evals/ejecucion_test.go`, `internal/evals/informe.go`, `internal/evals/informe_test.go`, `internal/evals/juez.go`, `internal/evals/juez_test.go`, `internal/evals/medida_test.go`, `internal/evals/sondeo_test.go`, `skills/boe-legislacion/SKILL.md`, `specs/020-h25-el-juez-de/plan.md`, `specs/020-h25-el-juez-de/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/020-h25-el-juez-de/gates/revision-cambios-r3.diff` (591 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
